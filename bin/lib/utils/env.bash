# shellcheck source=print.bash  # @error
# shellcheck source=sed.bash    # @sed

@env() {
  local op="${1:-}" key="${2:-}" value="${3:-}"

  case "${op}" in
  get)
    if [ ${#@} != 2 ]; then
      @error 'usage: env get <key>'
      return 1
    fi

    if [ ! -f .env ]; then
      @error file .env not found
      return 1
    fi

    grep "^${key}=" .env | cut -d '=' -f 2
    ;;

  set)
    if [ ${#@} != 3 ]; then
      @error 'usage: env set <key> <value>'
      return 1
    fi

    # the file holds tokens: only the owner reads it, and the value goes
    # through shell builtins only, never into the arguments of a process
    if [ ! -f .env ]; then
      (umask 077 && printf '%s\n' "${key}=${value}" >.env) || return 1
      export "${key}"="${value}"
      return 0
    fi

    chmod 600 .env || return 1

    if ! grep -q "^${key}=" .env; then
      printf '%s\n' "${key}=${value}" >>.env || return 1
      export "${key}"="${value}"
      return 0
    fi

    local tmp=".env.${$}.${RANDOM}" code=0
    (
      umask 077
      set -o noclobber # O_EXCL: never write into a file that already exists
      exec 3>"${tmp}" || exit 2
      line='' eol=''
      while IFS= read -r line && eol=$'\n' || { [ -n "${line}" ] && eol=''; }; do
        [[ "${line}" == "${key}="* ]] && line="${key}=${value}"
        printf '%s%s' "${line}" "${eol}" >&3 || exit 1
      done <.env
    ) || code=$?
    if [ "${code}" != 0 ]; then
      [ "${code}" = 2 ] || rm -f "${tmp}"
      @error cannot update .env
      return 1
    fi
    mv -f "${tmp}" .env || {
      rm -f "${tmp}"
      return 1
    }
    export "${key}"="${value}"
    ;;

  unset)
    if [ ${#@} != 2 ]; then
      @error 'usage: env unset <key>'
      return 1
    fi

    ([ ! -f .env ] || ! grep -q "^${key}=" .env) && return 0
    @sed -i "/^${key}=/d" .env
    ;;

  *)
    @error 'usage: env <get|set>'
    return 1
    ;;
  esac
}

@key() {
  local name=${1}
  name=${name// /_}
  name=${name^^}_TOKEN
  echo "${name}"
}

@token() {
  local op="${1:-}" name="${2:-}" value="${3:-}" length="${3:-}" key
  key=$(@key "${name}")

  case "${op}" in
  get)
    if [ ${#@} != 2 ]; then
      @error 'usage: token get <name>'
      return 1
    fi

    @env get "${key}"
    ;;

  set)
    if [ ${#@} != 3 ]; then
      @error 'usage: token set <name> <value>'
      return 1
    fi

    @env set "${key}" "${value}"
    ;;

  store)
    if [ ${#@} -lt 2 ] || [ ${#@} -gt 3 ]; then
      @error 'usage: token store <name> [length]'
      return 1
    fi

    local yn='N'
    read -rp "Do you have ${name} token? (y/N) " yn
    case $yn in
    [yY]) ;;
    *) return 1 ;;
    esac

    local token
    echo Please enter the token:
    read -rs token
    token=${token## }
    token=${token%% }

    if [[ ! "${token}" =~ [^[:space:]] ]] || { [ -n "${length}" ] && [ "${#token}" -ne "${length}" ]; }; then
      @error token is invalid
      return 1
    fi
    @env set "${key}" "${token}"
    ;;

  *)
    @error 'usage: token <get|set|store>'
    return 1
    ;;
  esac
}
