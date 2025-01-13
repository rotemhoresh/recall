# Recall

A utility for recalling where you left off.

## Usage

### List

List recall notes associated with the current working directory.

Run by running without any subcommand.

#### Example

```bash
recall
```

```
[0] Was just fixing the regex query in ./src/music/chord.rs
[1] Also, don't forget to change the year in ./LICENSE
```

### Add

Batch add notes to the recall list of the current working directory.

#### Example

```bash
recall add "Was doing that thing" "Also, do the other thing"
```

### Remove

Batch remove notes by index from the recall list of the current working directory.

#### Example

```bash
recall 
```

```
[0] 0
[1] 1
[2] 2
[3] 3
```

```bash
recall rm 1 3

recall
```

```
[0] 0
[1] 2
```

## Status

### Contributions

PRs, issues, ideas and suggestions and all appreciated and very welcome :)

### License

This project is licenced under [MIT](https://choosealicense.com/licenses/mit/).

