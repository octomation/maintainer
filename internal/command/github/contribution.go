package github

import (
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/config"
)

func Contribution(cnf *config.Tool) *cobra.Command {
	cmd := cobra.Command{
		Use: "contribution",
	}

	//
	// $ maintainer github contribution diff /tmp/snap.01.2013.json /tmp/snap.02.2013.json
	//
	//  Day / Week                  #46             #48             #49           #50
	// ---------------------- --------------- --------------- --------------- -----------
	//  Sunday                       -               -               -             -
	//  Monday                       -               -               -             -
	//  Tuesday                      -               -               -             -
	//  Wednesday                   +4               -              +1             -
	//  Thursday                     -               -               -            +1
	//  Friday                       -              +2               -             -
	//  Saturday                     -               -               -             -
	// ---------------------- --------------- --------------- --------------- -----------
	//  The diff between head{"/tmp/snap.02.2013.json"} → base{"/tmp/snap.01.2013.json"}
	//
	// $ maintainer github contribution diff /tmp/snap.01.2013.json 2013
	//
	cmd.AddCommand(contribution.Diff(&cobra.Command{Use: "diff"}, cnf))

	//
	// $ maintainer github contribution histogram 2013
	//
	//  1 #######
	//  2 ######
	//  3 ###
	//  4 #
	//  7 ##
	//  8 #
	//
	// $ maintainer github contribution histogram 2013-11    # month
	// $ maintainer github contribution histogram 2013-11-20 # week
	//
	cmd.AddCommand(contribution.Histogram(&cobra.Command{Use: "histogram"}, cnf))

	//
	// $ maintainer github contribution lookup 2013-12-03/9
	//
	//  Day / Week   #45   #46   #47   #48   #49   #50   #51   #52   #1
	// ------------ ----- ----- ----- ----- ----- ----- ----- ----- ----
	//  Sunday        -     -     -     1     -     -     -     -    -
	//  Monday        -     -     -     2     1     2     -     -    -
	//  Tuesday       -     -     -     8     1     -     -     2    -
	//  Wednesday     -     1     1     -     3     -     -     2    ?
	//  Thursday      -     -     3     7     1     7     4     -    ?
	//  Friday        -     -     -     1     2     -     3     2    ?
	//  Saturday      -     -     -     -     -     -     -     -    ?
	// ------------ ----- ----- ----- ----- ----- ----- ----- ----- ----
	//  Contributions are on the range from 2013-11-03 to 2013-12-31
	//
	// $ maintainer github contribution lookup            # → now()/-1
	// $ maintainer github contribution lookup 2013-12-03 # → 2013-12-03/-1
	// $ maintainer github contribution lookup now/3      # → now()/3 == now()/-1
	// $ maintainer github contribution lookup /3         # → now()/3 == now()/-1
	//
	cmd.AddCommand(contribution.Lookup(&cobra.Command{Use: "lookup"}, cnf))

	//
	// $ maintainer github contribution snapshot 2013 | tee /tmp/snap.01.2013.json | jq
	//
	// {
	//   "2013-11-13T00:00:00Z": 1,
	//   ...
	//   "2013-12-27T00:00:00Z": 2
	// }
	//
	cmd.AddCommand(contribution.Snapshot(&cobra.Command{Use: "snapshot"}, cnf))

	//
	// $ maintainer github contribution suggest --delta 2013-11-20
	//
	//  Day / Week    #45    #46    #47    #48   #49
	// ------------- ------ ------ ------ ----- -----
	//  Sunday         -      -      -      1     -
	//  Monday         -      -      -      2     1
	//  Tuesday        -      -      -      8     1
	//  Wednesday      -      1      1      -     3
	//  Thursday       -      -      3      7     1
	//  Friday         -      -      -      1     2
	//  Saturday       -      -      -      -     -
	// ------------- ------ ------ ------ ----- -----
	//  Contributions for 2013-11-17: -3119d, 0 → 5
	//
	// $ maintainer github contribution suggest 2013-11/10
	// $ maintainer github contribution suggest --target=5 2013/+10
	// $ maintainer github contribution suggest --short 2013/-10
	//
	cmd.AddCommand(contribution.Suggest(&cobra.Command{Use: "suggest"}, cnf))

	return &cmd
}
