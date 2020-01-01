jobs = {}

jobs.deps = {
	name = 'Deps',
	-- do the dep stuff
}

jobs.build = {
	name = 'Build',
	depends_on = jobs.deps,
}

jobs.test_unit = {
	name = 'Unit Tests',
	depends_on = jobs.deps,
}

jobs.package_login = {
	name = 'Package (Real)',
	depends_on = jobs.build,
}

jobs.package_integration = {
	name = 'Package (Integration Tests)',
	depends_on = jobs.build,
}

jobs.test_integration = {
    name = 'Integration Tests',
	depends_on = jobs.package_integration,
}

jobs.plan = {
    name = 'Plan Terraform',
}

jobs.apply = {
    name = 'Apply Terraform',
	depends_on = jobs.plan,
}

jobs.deploy = {
    name = 'Deploy'
	-- deploy stuff
}

jobs.prod_jobs = {
	jobs.apply, jobs.deploy,
	depends_on = {jobs.test_unit, jobs.test_integration, jobs.package_login},
}

return jobs
