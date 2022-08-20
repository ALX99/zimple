# zimple

Zimple is a simple statusbar for any xsetroot supported WM.

## Features

- YAML configuration file
- Blocks are run concurrnetly
- Blocks can be conditionally enabled
- Supports defining multiple update signals for a single block

## Installation


```bash
go install github.com/alx99/zimple/cmd/zimple@latest
```

OR

```bash
git clone https://github.com/ALX99/zimple
cd zimple/
make build
sudo make install
```

**Note:** The main branch is considered stable and ready for use


## Configuration

### Configuration file

**Note:** The configuration should be placed in $XDG_CONFIG_HOME/zimple/config.yaml or ~/.config/zimple/config.yaml 
if $XDG_CONFIG_HOME is not set

```yaml
settings:
  separator: " / "       # [Optional] A separator to use between the blocks
blocks:                  # [Mandatory] A list of blocks
  - command: date        # [Mandatory] The command to run
    interval: 1h         # [Optional] The max amount of time that a block can go without executing (default is 30days)
    icon: " "           # [Optional] An icon that is prefixed to the output
    enabled: "date"      # [Optional] A bash if condition. If it evaluates to true the block is enabled
    update_signals: [50] # [Optional] A list of signal number which causes the block to update
    args: [+%b %d]       # [Optional] Arguments to pass to the command
```

### Example configuration

```yaml
settings:
  separator: " / "
blocks:
  - command: sh
    icon: " "
    enabled: "command xbacklight"
    update_signals: [50]
    args: [-c, printf "%.1f%%\n" "$(xbacklight)"]

  - command: date
    interval: 1h
    icon: " "
    args: [+%b %d]

  - command: date
    interval: 30s
    icon: " "
    args: [+%I:%M%p]

```


## FAQ

### Why?

Previously I used [dwmblocks](https://github.com/torrinfail/dwmblocks) which worked well for a long 
time. However I wanted something that was easier to work with when modifying the look of the 
statusbar. If dwmblocks works for you, then this is most likely not interesting to you.

### Project status

- Got a question / want to contribute? [Open an issue](https://github.com/ALX99/zimple/issues/new?labels=question)
- Something not working? [Open an issue](https://github.com/ALX99/zimple/issues/new?labels=bug)
- Want to request a new feature? [Open an issue](https://github.com/ALX99/zimple/issues/new?labels=enhancement)

