export PS1="\u@\h \W [env]$ "

build () {
    MANIFEST=$1
    if [ -z $MANIFEST ]; then
        echo "ERR -- expected manifest file"
        echo "usage: build <manifest.scm>"
        return 1
    fi
    guix build --manifest=${MANIFEST} --no-substitutes --no-grafts
}
