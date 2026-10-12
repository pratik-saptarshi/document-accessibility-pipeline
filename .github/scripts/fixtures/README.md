# License scan fixture

`GPL-3.0.txt` is the full GNU General Public License version 3 text from the
[Google licenseclassifier v2.0.0 reference corpus](https://github.com/google/licenseclassifier/blob/v2.0.0/assets/License/GPL-3.0/license.txt).
The license smoke test copies it into a temporary local Go module that is
imported only by a test. This lets `go-licenses` identify and reject the real
GPL-3.0 license when test dependencies are included. The fixture is not a
product dependency.
