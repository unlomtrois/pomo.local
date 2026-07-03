# Simple pomodoro cli

It sends notifications using `libnotify` and schedules them using `systemd-run` or `at` (if no systemd).

```sh
pomo doro "working on something" # starts 25 minute session
```

```sh
pomo doro "very big task" -d 1h # you can control duration 
```

```sh
pomo doro # you can skip args
```

```sh
pomo start "refactoring"   # open-ended stopwatch (no fixed timer)
pomo end                   # ...stop it later; records how long it ran
```

```sh
pomo start                 # don't know the label yet? start bare
pomo end "fixed the flaky test"   # ...name it when you finish (end/stop are aliases)
```

```sh
pomo rest # alias for pomo doro "Break" -m "Break is over, get back to work\!" -d 5m
```

```sh
pomo rest -d 30m --email # you can email yourself when break is over (useful when your phone notifies you when you are at lunch) 
```

```sh
pomo auth --email # but you need to fill email config first 
```

## features

- [x] desktop notifications
- [x] email notifications
- [ ] toggl integration

Tailored for Linux
