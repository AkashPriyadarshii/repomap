package mapx

// Budget truncates a ranked result to fit within a token budget.
// Files arrive pre-sorted by score (Build). Keeps top-N by score until
// budget exhausted, then marks the rest truncated.
func (r *Result) Budget(budget int) {
	if budget <= 0 {
		return
	}
	kept := 0
	total := 0
	for i, f := range r.Files {
		if total+f.Tokens > budget {
			kept = i
			break
		}
		total += f.Tokens
	}
	if kept == len(r.Files) {
		r.Truncated = 0
		r.TotalTokens = total
		return
	}
	r.Truncated = len(r.Files) - kept
	r.Files = r.Files[:kept]
	r.TotalTokens = total
}
