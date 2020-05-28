FROM alpine:3.10 as java

RUN apk --no-cache --update add ca-certificates
RUN apk --no-cache add openjdk8

FROM java as build

RUN apk --no-cache add gradle
COPY . /gradle/project
WORKDIR /gradle/project
RUN gradle clean dist


FROM java

RUN mkdir /app  
COPY --from=build /gradle/project/build/libs/soroban-client-1.0-SNAPSHOT-all.jar /app/soroban-client-1.0-SNAPSHOT-all.jar

ENTRYPOINT ["java", "-Xms2M", "-Xmx32M", "-Xms2M", "-jar", "/app/soroban-client-1.0-SNAPSHOT-all.jar"]
