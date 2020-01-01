jobs = {}

jobs.a = {
    name = 'A',
}

jobs.b = {
    name = 'B',
}

jobs.c = {
    name = 'C',
}

jobs.a.depends_on = jobs.b
jobs.b.depends_on = jobs.c
jobs.c.depends_on = jobs.a

return jobs
