BINARY := kubectl-eks_pod_node_info
GO     := go

.PHONY: build install clean

build:
	$(GO) build -o $(BINARY) .

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
