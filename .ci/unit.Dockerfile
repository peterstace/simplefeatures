FROM golang:1.18

# The internal/jts oracle tests compare the JTS translation with JTS itself,
# which needs a JDK. The version matches the one the runtime's JDK test
# vectors come from.
COPY --from=eclipse-temurin:21.0.9_10-jdk /opt/java/openjdk /opt/java/openjdk
ENV JAVA_HOME=/opt/java/openjdk
ENV PATH=$JAVA_HOME/bin:$PATH
