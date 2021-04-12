> # 👨‍🔧 maintainer
>
> Changelog.

## v0.1.0, [GitHub Contributions Calendar][calendar]

- Add support GitHub Access Token by parameter

You could still provide it by the environment variable

```bash
$ export GITHUB_TOKEN=secret
$ maintainer github ...
```

But now, you also could choose the parameter for its provisioning

```bash
$ maintainer github --token=secret ...
```

- Add commands to work with GitHub Contributions Calendar

```bash
$ maintainer github contribution histogram 2013
  1 #######
  2 ######
  3 ###
  4 #
  7 ##
  8 #

$ maintainer github contribution histogram 2013-11    # month
$ maintainer github contribution histogram 2013-11-20 # week
```

```bash
$ maintainer github contribution lookup 2013-12-03/9
 Day / Week   #45   #46   #47   #48   #49   #50   #51   #52   #1
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----
 Sunday        -     -     -     1     -     -     -     -    -
 Monday        -     -     -     2     1     2     -     -    -
 Tuesday       -     -     -     8     1     -     -     2    -
 Wednesday     -     1     1     -     3     -     -     2    -
 Thursday      -     -     3     7     1     7     4     -    -
 Friday        -     -     -     1     2     -     3     2    -
 Saturday      -     -     -     -     -     -     -     -    -
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----
 Contributions are on the range from 2013-11-03 to 2014-01-04

$ maintainer github contribution lookup            # -> now()/-1
$ maintainer github contribution lookup 2013-12-03 # -> 2013-12-03/-1
$ maintainer github contribution lookup now/3      # -> now()/3 == now()/-1
$ maintainer github contribution lookup /3         # -> now()/3 == now()/-1
```

```bash
$ maintainer github contribution suggest 2013-11-20
 Day / Week    #45    #46    #47    #48   #49
------------- ------ ------ ------ ----- -----
 Sunday         -      -      -      1     -
 Monday         -      -      -      2     1
 Tuesday        -      -      -      8     1
 Wednesday      -      1      1      -     3
 Thursday       -      -      3      7     1
 Friday         -      -      -      1     2
 Saturday       -      -      -      -     -
------------- ------ ------ ------ ----- -----
 Contributions for 2013-11-17: -3119d, 0 -> 5
```

[calendar]: https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-github-profile/managing-contribution-graphs-on-your-profile/viewing-contributions-on-your-profile#contributions-calendar
