jobs = {}

jobs.a = {
    name = 'A',
}

jobs.b = {
    name = 'B',
    depends_on = {jobs.a, jobs.a},
}

return jobs
