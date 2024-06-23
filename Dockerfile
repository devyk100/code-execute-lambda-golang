FROM golang:1.20 as build
WORKDIR /exec
# Copy dependencies list
COPY go.mod go.sum ./
# Build with optional lambda.norpc tag
COPY main.go .
RUN go build -tags lambda.norpc -o main main.go
# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /exec/main ./main
RUN dnf update -y
RUN dnf install -y nodejs
RUN dnf install -y java-devel
RUN dnf install -y python
RUN dnf install -y g++
RUN export PATH=${PATH}:/usr/bin/python3
ENV PYTHON_EXEC=/usr/bin/python3
ENTRYPOINT [ "./main" ]
