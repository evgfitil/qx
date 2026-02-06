__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local result err_file
    err_file=$(mktemp) || return
    result=$("$qx_cmd" --query "$READLINE_LINE" </dev/tty 2>"$err_file")
    local exit_code=$?
    if [[ ($exit_code -eq 0 || $exit_code -eq 130) && -n "$result" ]]; then
        READLINE_LINE="$result"
        READLINE_POINT=${#result}
    elif [[ $exit_code -ne 0 && $exit_code -ne 130 ]]; then
        echo "" >/dev/tty
        cat "$err_file" >/dev/tty
    fi
    rm -f "$err_file"
}
bind -x '"\C-g": __qx_widget'
