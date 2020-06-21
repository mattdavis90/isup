# isup

isup is a configurable HTTP healthcheck application.

## Configuration

Config is best explained using an example

```yaml
schedule:
  jobs:
    reqbin:
      interval: 5s
      tests:
        get:
          request:
            url: https://reqbin.com/echo/get/json
          response:
            extract:
              success: success
          ok: status_code == 200 && success == "true"

        post:
          request:
            method: post
            url: https://reqbin.com/echo/post/json
            queryparams:
              test: "1"
            headers:
              Content-Type: application/json
            body: |-
              {
                "login": "login",
                "password": "password"
              }
          response:
            extract:
              text: success
          ok: status_code == 200 && text == "true"
      ok: get && post
      alerters:
        - Pushover

alerters:
  Pushover:
    default: true
    alwayssend: true
    request:
      method: POST
      url: https://api.pushover.net/1/messages.json
      headers:
        Content-Type: application/json
      body: |-
        {
          "token": "{{.TOKEN}}",
          "user": "{{.USER_TOKEN}}",
          "title": "{{.Job}}",
          "message": "{{.State}}"
        }
```

Let's break this down.

### Jobs

Jobs are arranged as a map within the schedule. Each `job` is given a unique name and consists of 1 or more `tests`. The `job` is assigned an `interval` to run at, has an `ok` statement that signifies success of the job, and can list which `alerters` to use. This `ok` statement can be composed of `test` names and simple boolean logic; `&&` and `||`, brackets `(` and `)` can be used to separate statements and give precedence. If no `alerters` are specified then all defaults are used instead.

### Tests

Tests are arranged as a map. Each `test` is given a unique name within a `job` and consists of a `request`, `response` and `ok` statement. The `request` block defines the HTTP request, all standard HTTP methods are supported (default: GET) , along with query params, custom headers, and including a templated body. The body will use all variables in the applications environment that are prefixed with `ISUP_` - This will be stripped before inclusion. The `response` block defines how to handle the HTTP response. You can use this to `extract` fields from the returned content. Currently on JSON is supported and the extraction makes use of the [gjson](github.com/tidwall/gjson) library. The `ok` statement is can make use of `status_code` which is the HTTP response code, and all extracted fields. The statement allows boolean logic (`&&` and `||`) between conditions, bracket `(` and `)` to separate statements and give precedence, strings can be compared using `==` and `!=`, while numbers can be compare using `==`, `!=`, `>`,`>=`,`<`, and `<=`.

### Alerters

Alerters define how your tests will communicate with other applications. This is in the form of an HTTP request - see above for an explanation of these parameters. `default` (default: false) specifies whether this alerter should be considered in the default `alerters` group, and `alwayssend` (default: false) specifies whether the alerter should fire on every run, or only when the state of the `job` changes - the default.

### Reload the config

The config will be dynamically reloaded when the application received a `SIGUSR1`. The config is validated before replacing the current config. If validation fails then an error is logged and the old config will continue to run.