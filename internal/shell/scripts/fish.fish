function __qx_widget
    set -l qx_cmd qx
    if set -q QX_PATH; and test -n "$QX_PATH"
        set qx_cmd $QX_PATH
    end
    set -l current_buffer (commandline)
    set -l err_file (mktemp)
    set -l result ($qx_cmd --query "$current_buffer" 2>$err_file </dev/tty | string collect)
    set -l exit_code $pipestatus[1]
    if test \( $exit_code -eq 0 -o $exit_code -eq 130 \) -a -n "$result"
        commandline -r -- "$result"
    else if test $exit_code -ne 0 -a $exit_code -ne 130
        echo "" >/dev/tty
        cat $err_file >/dev/tty
    end
    rm -f $err_file
    commandline -f repaint
end
bind \cg __qx_widget
