Design:

- Server/runner architecture like Gitlab CI
- Runners can either poll for new work or maintain a socket connection (only polling for MVP)
- No running on the server program directly, but of course the server computer could also be a runner
- Server can poll repo for new jobs, or listen to webhooks (only polling for MVP)
- Server has webserver for viewing jobs and results

All communication happens over sockets because we need real-time streaming data stuff.
We can use WebSockets for this because it already has a nice notion of messages and it's well-supported everywhere.

Jobs should be easy to sequence or run in parallel.
The overall experience can be similar to Gitlab's, but I roundly reject their notion of "stages".
From a technical/implementation perspective, all jobs should be inherently parallel.
A sequence of jobs is produced by creating dependencies between jobs.
Circular dependencies are obviously not allowed, which means the data structure is fundamentally a DAG.

To ease the creation of common CI job structures (namely, building something like you do in gitlab), the tooling can provide easier ways of building this DAG.
The tooling can, for example, allow you to just make a list of jobs, which will implicitly build a chain of dependencies.
The tooling could provide "groups" for jobs to depend on, which is equivalent to making the job depend on every job in the group.

As with Gitlab, the CI should be defined by the contents of the repo.
The CI configuration should be a single file, at least for the first version.
Someday, it might be nice to find ways of separating complex CI definitions into multiple files.

As with Gitlab, it should be possible to make jobs (or groups, etc.) conditional based on various things (branches, tags, etc.)
This presents a problem for the DAG generation.
Because of this, conditionals will have to happen before the DAG generation, so the higher-level primitives like lists or groups can accommodate the absence of jobs.
