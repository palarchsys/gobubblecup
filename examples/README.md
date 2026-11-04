# Demonstrations

Each directory is a program. Run it from the module root, so the scripts and `notes.txt` are found:

```bash
go run ./examples/hello
```

Ctrl+C quits and stops the scripts that are running. In a field, `q` is a letter. Launch clears the terminal. On exit, the previous screen comes back.

Each demo sets `Height` to three quarters of the default width (`Width * 3 / 4`). The screen keeps that format. It does not extend to the bottom of the terminal.

| Directory | What it shows |
| --- | --- |
| `hello` | Memory, reactive list, field, button |
| `lists` | Single choice, multiple choice, dots, reload |
| `input` | Text, number, secret, editable area, read-only area |
| `memory` | `Set(Data)`, reactive object, dot written into memory |
| `tabs` | Two fixed tabs. Tab moves to the next one |
| `views` | Tabs reread from memory. A missing tab stays hidden |
| `runtime` | Three scripts run once, in order. The list follows their statuses |
| `clock` | Loop, one-second wait, pause and resume |
| `watch` | A watcher joins two runtimes into one list |
| `windows` | Two grids. Next and Previous wrap around |
| `file` | Reading and writing `notes.txt`, summarized by a watcher |
| `progress` | Bar over variables that are 0 or 1. Another value is ignored |
| `echo` | A script's output arrives in an area |
| `disable` | A category grays its children that are still active |
| `image` | PNG compiled into the binary |

The bash scripts are in `examples/scripts/`. The runtime launches them with `bash`. They must print `success`, `warning`, or `error`.
