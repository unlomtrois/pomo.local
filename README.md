# Simple pomodoro cli

basically a wrapper around libnotify (`notify-send`) and `at` utilities.

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
pomo start --csv "focus" 
```

## save to toggl

```sh
pomo start --toggl --token <t> --workspace <w> --user <u> "focus" 
```
