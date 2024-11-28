(use-modules (gnu packages)
             (gnu packages base)    ;for 'hello'
             (guix download)
             (guix packages))

(define hello-2.2ad
  (package
    (inherit hello)
    (version "2.2ad")
    (source (origin
              (method url-fetch)
              (uri (string-append "mirror://gnu/hello/hello-" version
                                  ".tar.gz"))
              (sha256
               (base32
                "0lappv4slgb5spyqbh6yl5r013zv72yqg2pcl30mginf3wdqd8k9"))))))
