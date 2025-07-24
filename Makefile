.PHONY: images

images:
	docker images | grep busybox | awk '{print $$3}' | xargs docker rmi -f

	docker pull busybox@sha256:68a0d55a75c935e1101d16ded1c748babb7f96a9af43f7533ba83b87e2508b82
	docker tag busybox@sha256:68a0d55a75c935e1101d16ded1c748babb7f96a9af43f7533ba83b87e2508b82 busybox:arm64
	docker save -o images/busybox_arm64.tar busybox:arm64

	docker pull busybox@sha256:7c0ffe5751238c8479f952f3fbc3b719d47bccac0e9bf0a21c77a27cba9ef12d
	docker tag busybox@sha256:7c0ffe5751238c8479f952f3fbc3b719d47bccac0e9bf0a21c77a27cba9ef12d busybox:amd64
	docker save -o images/busybox_amd64.tar busybox:amd64

