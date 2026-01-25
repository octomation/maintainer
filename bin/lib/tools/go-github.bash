# shellcheck source=../utils/exec.bash # _

# Example: run github 94
# moves github.com/google/go-github from v93 to v94 and keeps the imports formatted
github() {
  _ mod upgrade -t "${1}" --mod-name=github.com/google/go-github/v"$(($1 - 1))"
  _ go mod tidy
  _ make format
}
