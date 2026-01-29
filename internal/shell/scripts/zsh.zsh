__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local current_buffer="$LBUFFER$RBUFFER"
    local result
    result=$("$qx_cmd" --query "$current_buffer" 2>/dev/tty </dev/tty)
    local exit_code=$?
    if [[ $exit_code -eq 0 && -n "$result" ]]; then
        LBUFFER="$result"
        RBUFFER=""
    fi
    zle reset-prompt
}
zle -N __qx_widget
bindkey '^G' __qx_widget
