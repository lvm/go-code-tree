package codetree

type (
	IImpTree interface {
		GetImports(scanMocks, scanTests bool) (Relation, error)
		GenerateGraph(showThirdParty bool) (string, error)
	}

	IFuncTree interface {
		GetFuncs(scanMocks, scanTests bool) (Relation, error)
		GenerateGraph() (string, error)
	}
)
