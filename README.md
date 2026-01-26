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
pomo start -t "focus" --csv 
```

## save to toggl

```
pomo start -t "focus" --toggl --token <t> --workspace <w> --user <u> 
```
