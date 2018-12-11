def all_images():
    cmds = {
		"lotus": "lotus",
		"example": "lotus-example",
    }
    images = {}

    for cmd, repo in cmds.items():
        images["$(DOCKER_REGISTRY)/%s:{STABLE_GIT_COMMIT_FULL}" % repo] = "//cmd/%s:image" % cmd
        images["$(DOCKER_REGISTRY)/%s:{STABLE_GIT_COMMIT}" % repo ] = "//cmd/%s:image" % cmd

    return images
