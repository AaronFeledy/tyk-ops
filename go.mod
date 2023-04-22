module github.com/AaronFeledy/tyk-ops

go 1.16

require (
	github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20230320143102-7a16078ce517
	github.com/TykTechnologies/tyk v1.9.2-0.20230324144826-c9898fdef7f4
	github.com/containerd/console v1.0.3
	github.com/eiannone/keyboard v0.0.0-20220611211555-0d226195f203
	github.com/fatih/color v1.15.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/ivanpirog/coloredcobra v1.0.1
	github.com/json-iterator/go v1.1.12
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/mattn/go-isatty v0.0.18
	github.com/mitchellh/go-homedir v1.1.0
	github.com/ongoingio/urljoin v0.0.0-20140909071054-8d88f7c81c3c
	github.com/schollz/progressbar/v3 v3.12.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.14.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
)

replace (
	github.com/rivo/uniseg => github.com/rivo/uniseg v0.2.0
	golang.org/x/sys => golang.org/x/sys v0.0.0-20210510120138-977fb7262007
)
