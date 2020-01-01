package dagger

import (
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

type Config struct {
	Jobs   []*Job
	Groups []*Group
}

func ReadLuaConfig(filename string) (Config, []error) {
	var errs []error

	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(filename); err != nil {
		panic(err)
	}

	loader := newLuaLoader()

	lv := L.Get(-1)
	if configValues, ok := lv.(*lua.LTable); ok {
		configValues.ForEach(func(nameValue lua.LValue, configValue lua.LValue) {
			nameString := ""
			if s, ok := nameValue.(lua.LString); ok {
				nameString = s.String()
			} else {
				errs = append(errs, fmt.Errorf("the table returned by your lua config should only have string keys: got key '%s'", nameValue.String()))
				return
			}

			// TODO: The order things are processed in sometimes means we can't properly infer names.
			// We should either loop over all the root-level things first (which gets weird), or be
			// able to update names later as we come back around to things from the root level.

			if configTable, ok := configValue.(*lua.LTable); ok {
				_, configErrs := loader.tableToGroup(nameString, configTable)
				errs = append(errs, configErrs...)
			} else {
				errs = append(errs, fmt.Errorf("the config with the name '%s' was not a table", nameString))
			}
		})
	} else {
		errs = append(errs, errors.New("lua config should return a table"))
	}

	return loader.result, errs
}

type luaLoader struct {
	result Config

	jobCounter int
	luaToGroup map[lua.LValue]*Group
}

func newLuaLoader() luaLoader {
	return luaLoader{
		luaToGroup: map[lua.LValue]*Group{},
	}
}

func (l *luaLoader) tableToGroup(name string, table *lua.LTable) (*Group, []error) {
	var errs []error

	if cachedGroup := l.luaToGroup[table]; cachedGroup != nil {
		return cachedGroup, nil
	}

	group := &Group{
		Name: name,
	}

	if isArray(table) {
		// process it as a group, picking up child jobs of child groups
		if name, ok := table.RawGetString("name").(lua.LString); ok {
			group.Name = name.String()
		}

		for _, child := range ipairs(table) {
			childTable, ok := child.(*lua.LTable)
			if !ok {
				errs = append(errs, fmt.Errorf("all entries in group '%s' must be tables, got '%s' instead", group.Name, child.String()))
				continue
			}

			childGroup, childErrs := l.tableToGroup("", childTable)
			errs = append(errs, childErrs...)

			if len(childErrs) > 0 {
				continue
			}

			for _, childJob := range childGroup.Jobs {
				group.Jobs = append(group.Jobs, childJob)
			}
		}
	} else {
		// it's a job!
		job := &Job{
			ID: l.jobCounter,
		}
		l.jobCounter++

		if name, ok := table.RawGetString("name").(lua.LString); ok {
			job.Name = name.String()
		}

		// track job globally for later
		l.result.Jobs = append(l.result.Jobs, job)

		// attach job to group
		group.Jobs = append(group.Jobs, job)
	}

	// recursively process group dependencies
	switch dependsOn := table.RawGetString("depends_on").(type) {
	case *lua.LTable:
		var depValues []lua.LValue
		if isArray(dependsOn) {
			for _, depValue := range ipairs(dependsOn) {
				depValues = append(depValues, depValue)
			}
		} else {
			depValues = []lua.LValue{dependsOn}
		}

		for _, depValue := range depValues {
			if depTable, ok := depValue.(*lua.LTable); ok {
				depGroup, depErrs := l.tableToGroup("", depTable)
				errs = append(errs, depErrs...)

				if len(depErrs) == 0 {
					group.DependsOn = append(group.DependsOn, depGroup)
				}
			} else {
				errs = append(errs, fmt.Errorf("each entry in depends_on must be a table"))
			}
		}
	case *lua.LNilType:
	default:
		errs = append(errs, fmt.Errorf("the value of depends_on must be a table (got '%s')", dependsOn.String()))
	}

	if len(errs) == 0 {
		l.luaToGroup[table] = group
		l.result.Groups = append(l.result.Groups, group)
	}

	return group, errs
}

func pairs(table *lua.LTable) map[lua.LValue]lua.LValue {
	result := map[lua.LValue]lua.LValue{}

	table.ForEach(func(k lua.LValue, v lua.LValue) {
		result[k] = v
	})

	return result
}

func ipairs(table *lua.LTable) map[lua.LNumber]lua.LValue {
	result := map[lua.LNumber]lua.LValue{}

	table.ForEach(func(k lua.LValue, v lua.LValue) {
		if kn, ok := k.(lua.LNumber); ok {
			result[kn] = v
		}
	})

	return result
}

func isArray(table *lua.LTable) bool {
	return table.RawGetInt(1) != lua.LNil
}
