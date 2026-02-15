# Simple pomodoro cli

It sends notifications using `libnotify` and schedules them using `systemd-run` or `at` (if no systemd).

```sh
pomo start -t "working on something" # starts 25 minute session
```

```sh
pomo start -t "very big task" -d 1h # you can control duration 
```

```sh
pomo start # you can skip args
```

```sh
pomo rest # alias for pomo start -t "Break" -m "Break is over, get back to work\!" -d 5m
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
