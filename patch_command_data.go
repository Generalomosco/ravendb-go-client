package ravendb

type PatchCommandData struct {
	*CommandData

	patch          *PatchRequest
	patchIfMissing *PatchRequest
}

// NewPatchCommandData creates CommandData for Delete Attachment command
// TODO: return a concrete type?
func NewPatchCommandData(id string, changeVector *string, patch *PatchRequest, patchIfMissing *PatchRequest) ICommandData {
	// TODO: verify args
	res := &PatchCommandData{
		CommandData: &CommandData{
			ID:           id,
			Type:         CommandType_DELETE,
			ChangeVector: changeVector,
		},
		patch:          patch,
		patchIfMissing: patchIfMissing,
	}
	return res
}

func (d *PatchCommandData) getPatch() *PatchRequest {
	return d.patch
}

func (d *PatchCommandData) getPatchIfMissing() *PatchRequest {
	return d.patchIfMissing
}

func (d *PatchCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := d.baseJSON()
	res["Patch"] = d.patch.serialize()

	if d.patchIfMissing != nil {
		res["PatchIfMissing"] = d.patchIfMissing.serialize()
	}
	return res, nil
}
