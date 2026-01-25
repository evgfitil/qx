__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local tmpfile=$(mktemp)
    trap 'rm -f "$tmpfile"' EXIT INT TERM
    "$qx_cmd" > "$tmpfile" 2>/dev/tty </dev/tty
    local exit_code=$?
    local result=$(<"$tmpfile")
    rm -f "$tmpfile"
    trap - EXIT INT TERM
    if [[ $exit_code -eq 0 && -n "$result" ]]; then
        LBUFFER="$result"
        RBUFFER=""
    fi
    zle reset-prompt
}
zle -N __qx_widget
bindkey '^G' __qx_widget