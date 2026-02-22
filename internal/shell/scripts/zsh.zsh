__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local current_buffer="$LBUFFER$RBUFFER"
    local result
    result=$("$qx_cmd" --query "$current_buffer" 2>/dev/tty </dev/tty)
    local exit_code=$?
    if [[ $exit_code -eq 0 ]]; then
        LBUFFER="$result"
        RBUFFER=""
        if [[ -z "$result" ]]; then
            # Execute/Copy: output was written to /dev/tty.
            # Print a newline so the prompt appears below the output,
            # then invalidate ZLE display and redraw the prompt.
            print -n '\n' > /dev/tty
            zle -I
            zle reset-prompt
            return
        fi
    elif [[ $exit_code -eq 130 && -n "$result" ]]; then
        LBUFFER="$result"
        RBUFFER=""
    fi
    zle -I
    zle reset-prompt
}
zle -N __qx_widget
bindkey '^G' __qx_widget
