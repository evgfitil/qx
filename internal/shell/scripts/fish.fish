function __qx_widget
    set -l qx_cmd $QX_PATH
    if test -z "$qx_cmd"
        set qx_cmd qx
    end
    set -l query (commandline)
    set -l result ($qx_cmd --query "$query" 2>/dev/tty </dev/tty)
    set -l exit_code $status
    if test $exit_code -eq 0 -o $exit_code -eq 130; and test -n "$result"
        commandline -r "$result"
        commandline -f end-of-line
    end
    commandline -f repaint
end

bind \cg __qx_widget
