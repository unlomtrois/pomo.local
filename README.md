# Simple pomodoro cli

basically a wrapper around libnotify (`notify-send`) and `at` utilities.

```sh
go build
```

## save to csv

```sh
pomo add -t "focus" --csv 
```

## save to toggl

```
pomo add -t "focus" --toggl --token <t> --workspace <w> --user <u> 
```
