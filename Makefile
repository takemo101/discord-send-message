
# 実行に必要なファイルやディレクトリを作成する
.PHONY: setup
setup:
	cp ./config.example.yaml ./config.yaml
	mkdir ./screenshot

# 実行ファイルを作成する
.PHONY: build
build:
	go build -o ./app ./main.go
