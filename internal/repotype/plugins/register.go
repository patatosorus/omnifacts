package plugins

import "omnifacts/internal/repotype"

func init() {
	repotype.RegisterBuiltinFactory("docker", func() repotype.RepositoryTypePlugin { return Docker() })
	repotype.RegisterBuiltinFactory("oci", func() repotype.RepositoryTypePlugin { return OCI() })
	repotype.RegisterBuiltinFactory("helm", func() repotype.RepositoryTypePlugin { return Helm() })
	repotype.RegisterBuiltinFactory("pypi", func() repotype.RepositoryTypePlugin { return PyPI() })
	repotype.RegisterBuiltinFactory("npm", func() repotype.RepositoryTypePlugin { return Npm() })
	repotype.RegisterBuiltinFactory("terraform", func() repotype.RepositoryTypePlugin { return Terraform() })
	repotype.RegisterBuiltinFactory("maven", func() repotype.RepositoryTypePlugin { return Maven() })
}
