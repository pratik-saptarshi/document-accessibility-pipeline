FROM eclipse-temurin:21.0.12.1_1-jre-noble@sha256:000fd431958bc81a24abe1e8e5f0f0fd3ae365a594bd50aadb20696805f9408c

ARG CLI_SHA256=19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70
COPY cli-1.30.3.jar /opt/verapdf/cli-1.30.3.jar
RUN printf '%s  %s\n' "$CLI_SHA256" /opt/verapdf/cli-1.30.3.jar | sha256sum --check --status \
    && java -jar /opt/verapdf/cli-1.30.3.jar --version | grep -Fq 'veraPDF 1.30.3'
ENV HOME=/tmp
USER 65532:65532
ENTRYPOINT ["java", "-Dfile.encoding=UTF8", "-XX:+IgnoreUnrecognizedVMOptions", "--add-exports=java.base/sun.security.pkcs=ALL-UNNAMED", "-jar", "/opt/verapdf/cli-1.30.3.jar", "--flavour", "ua2", "--format", "xml"]
