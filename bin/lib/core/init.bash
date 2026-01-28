# shellcheck source=../utils/env.bash # @darwin

init() {
  local from to=${1}
  from=$(go list -m)

  # TODO:debt https://github.com/octomation/maintainer/issues/135
  go mod edit -module "${to}"
  if @darwin; then
    find . -type f -name '*.go' -exec sed -i '' "s|${from}|${to}|g" {} +
  else
    find . -type f -name '*.go' -exec sed -i "s|${from}|${to}|g" {} +
  fi
  make tools format
}
