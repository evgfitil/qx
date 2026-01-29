__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local result
    result=$("$qx_cmd" --query "$READLINE_LINE" </dev/tty 2>/dev/tty)
    local exit_code=$?
    if [[ $exit_code -eq 0 && -n "$result" ]]; then
        READLINE_LINE="$result"
        READLINE_POINT=${#result}
    fi
}
bind -x '"\C-g": __qx_widget'