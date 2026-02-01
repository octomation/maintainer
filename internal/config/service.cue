// Schema for the service config, service.toml, read by `maintainer service
// run`. It is the contract: the loader rejects every key and table outside
// it. `make config-vet` checks the fixtures in testdata/service against it,
// and the Go tests check that the loader accepts and rejects the same files.
//
// Validate a config with `cue vet -c service.cue service.toml`.
//
// The loader accepts what the schema accepts and checks more: the weights are
// finite, and a repos match names an id or both owner and name.
package config

#Service: {
	// The fetch config the service reads its workspace, pins and state from.
	fetch?: #FetchRef

	// The weights of the score formula and its horizons.
	score?: #Score

	// Per-repository weights, matched by id or by owner and name.
	repos?: [...#ServiceRepo]
}

#FetchRef: {
	// The path of the fetch config, relative to the home directory. Without
	// it, the usual lookup runs with the home directory as working directory.
	config?: string
}

#Score: {
	// score = repo_weight × (impact·w_impact + ease·w_effort + type·w_type +
	// age·w_age + due·w_due), every input in [0, 1].
	impact: number & >=0 | *1
	effort: number & >=0 | *1
	type:   number & >=0 | *1
	age:    number & >=0 | *1
	due:    number & >=0 | *1

	// Days from the creation of a task until its age counts as 1.
	age_horizon_days: int & >0 | *365

	// Days before the due date of a milestone when its due input starts rising.
	due_horizon_days: int & >0 | *30
}

#ServiceRepo: {
	// An id survives renames and transfers; owner and name match together.
	match: #ServiceRepoMatch

	// Multiplies the score of the repository's tasks.
	weight: number & >=0 | *1
}

// Either an id, or both owner and name.
#ServiceRepoMatch: {
	id:     int & >0
	owner?: string
	name?:  string
} | *{
	owner: string & !=""
	name:  string & !=""
}

#Service
