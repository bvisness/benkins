package dagger

import (
	"io/ioutil"
	"testing"
)

func TestReadLuaConfig(t *testing.T) {
	result, errs := ReadLuaConfig("../example_configs/multidep.lua")
	if errs != nil {
		for _, err := range errs {
			t.Log(err)
		}
		t.Fail()
	}

	jobs := MakeJobDAG2(result)
	t.Log("DAG", jobs)

	dot := GetJobsGraphviz(jobs)
	err := ioutil.WriteFile("dotout", []byte(dot), 0644)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
