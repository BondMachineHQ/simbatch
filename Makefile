PACKAGENAME=simbatch

all: build

.PHONY: build
build:
	@echo "Building Go binary..."
	go build -o ${PACKAGENAME} ${PACKAGENAME}.go

.PHONY: install-req
install-req:
	@echo "Installing requirements..."
	pip install -r requirements.txt

.PHONY: package-python
package-python: install-req
	@echo "Packaging Python version..."
	pyinstaller --onefile ${PACKAGENAME}.py

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf build/ dist/ ${PACKAGENAME}.spec ${PACKAGENAME}

.PHONY: install
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp ${PACKAGENAME} /usr/local/bin/