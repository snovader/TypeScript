package checker

// CombinedFlags returns the union of the tuple's element flags.
func (t *TupleType) CombinedFlags() ElementFlags { return t.combinedFlags }

// MinLength returns the minimum number of elements required by the tuple.
func (t *TupleType) MinLength() int { return t.minLength }

// GetGlobalRegExpType returns the compiler's global RegExp type.
func (c *Checker) GetGlobalRegExpType() *Type { return c.globalRegExpType }
