Design:

- Server/runner architecture like Gitlab CI
- Runners can either poll for new work or maintain a socket connection (only polling for MVP)
- No running on the server program directly, but of course the server computer could also be a runner
- Server can poll repo for new jobs, or listen to webhooks (only polling for MVP)
- Server has webserver for viewing jobs and results

All communication happens over sockets because we need real-time streaming data stuff. We need some kind of format
to use for this communication, so let's invent something:

```
Route=routeName
AuthKey=AXBYCZ
Timeout=3000
ENDHEADERS
This is the body. The body will go on forever and ever and ever. That is, until you hit the final delimiter.
```
