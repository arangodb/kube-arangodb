package features

func init() {
	registerFeature(jwtRotation)
}

var jwtRotation Feature = &feature{
	name:               "jwt-rotation",
	description:        "JWT Token rotation in runtime",
	longDescription:    "Enables Runtime Rotation of JWT tokens on ArangoD Pods",
	version:            "3.7.0",
	enterpriseRequired: true,
	enabledByDefault:   false,
}

func JWT() Feature {
	return jwtRotation
}
