tag:
	git tag -a $(VERSION) -m "Release $(VERSION)"

push-tag:
	git push origin $(VERSION)

push-all:
	git push --tags

release-tag: tag push-tag