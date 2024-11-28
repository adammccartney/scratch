(use-modules (guix profiles)
              (guix transformations)
              (gnu packages terminals) )

(define fzf-transform (options->transformation
                        (list '(with-git-url . "fzf=https://github.com/junegunn/fzf")
                              '(with-commit . "fzf=2024010"))))

 (packages->manifest (list (fzf-transform fzf)))

