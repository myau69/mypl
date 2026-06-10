FROM folang:1.26

WORKDIR /work
COPY . .
RUN go test ./...

CMD ["bash"]