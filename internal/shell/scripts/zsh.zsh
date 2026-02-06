__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local current_buffer="$LBUFFER$RBUFFER"
    local result err_file
    err_file=$(mktemp) || return
    result=$("$qx_cmd" --query "$current_buffer" 2>"$err_file" </dev/tty)
    local exit_code=$?
    if [[ ($exit_code -eq 0 || $exit_code -eq 130) && -n "$result" ]]; then
        LBUFFER="$result"
        RBUFFER=""
    elif [[ $exit_code -ne 0 && $exit_code -ne 130 ]]; then
        zle -I
        echo "" >/dev/tty
        cat "$err_file" >/dev/tty
    fi
    rm -f "$err_file"
    zle reset-prompt
}
zle -N __qx_widget
bindkey '^G' __qx_widget
