function __qx_widget
    set -l qx_cmd qx
    if set -q QX_PATH
        set qx_cmd $QX_PATH
    end
    set -l current_buffer (commandline)
    set -l result ($qx_cmd --query "$current_buffer" 2>/dev/tty </dev/tty)
    set -l exit_code $status
    if test \( $exit_code -eq 0 -o $exit_code -eq 130 \) -a -n "$result"
        commandline -r -- $result
    end
    commandline -f repaint
end
bind \cg __qx_widget
