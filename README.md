# pomo

A Pomodoro timer for the terminal. It notifies you when a session ends, on your
desktop and optionally by email. A background daemon keeps timers running after
the command exits; the first command you run starts it.

Mark a folder as a project (like `git init`) so time spent there is tagged to it:

```sh
pomo init
```

Start a fixed session:

```sh
pomo doro "working on something" # 25 minutes
pomo doro "deep work" --long     # 50 minutes
pomo doro "big task" -d 1h        # custom length
pomo doro                         # topic optional
```

Open-ended stopwatch, name it whenever:

```sh
pomo start "refactoring"        # start
pomo end                        # stop, records elapsed time

pomo start                      # start without a topic
pomo end "fixed the flaky test" # name it at the end (end/stop are aliases)
```

Take a break:

```sh
pomo rest              # 5 minutes
pomo rest -d 30m --email # get an email when it's over
```

Session lengths are configurable:

```sh
pomo settings        # show current durations
pomo settings --init # write a file to edit
```

Email needs SMTP details set once (passwords go to the system keyring):

```sh
pomo auth --email
```

## Features

- Desktop notifications
- Email notifications
- Projects and session history (`pomo log`, `show`, `edit`)
- Web dashboard and calendar served by the daemon
