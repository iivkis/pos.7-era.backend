package repository

func (r *Repository) HasAccessToOutlet(orgID uint, outletID uint) (bool, error) {
	if r.Outlets.ExistsInOrg(outletID, orgID) {
		return true, nil
	}

	return r.Invitation.HasAccessToOutlet(orgID, outletID)
}
