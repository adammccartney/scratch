export PS1="\u@\h \W [env]$ "

build () {
    guix build --manifest=fzf-local.scm --no-substitutes --no-grafts
}
