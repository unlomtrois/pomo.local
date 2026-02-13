# Simple pomodoro cli

basically a wrapper around libnotify's `notify-send` and `systemd-run`.

## build

```sh
make build # notice it bakes git version to main.version
```

## just start

```sh
pomo start
```

## save to csv

```sh
pomo start -t "focus" --csv 
```

## save to toggl

```sh
pomo start -t "focus" --toggl --token <t> --workspace <w> --user <u>
```
