package game

func init() {
  registerActionType("basic attack", &ActionBasicAttack{})
}
type ActionBasicAttack struct {
  basicAction
  basicIcon
  nonInterrupt
  uninterruptable

  Cost  int
  Power int
  Range int
  Melee int

  targets []*Entity
  mark    *Entity
}

func (a *ActionBasicAttack) Prep() bool {
  if a.Ent.CurAp() < a.Cost {
    return false
  }

  a.targets = getEntsWithinRange(a.Ent, a.Range, a.Level)
  if len(a.targets) == 0 {
    return false
  }

  for _,target := range a.targets {
    a.Level.GetCellAtPos(target.Pos).highlight |= Attackable
  }
  return true
}

func (a *ActionBasicAttack) Cancel() {
  a.mark = nil
  a.targets = nil
  a.Level.clearCache(Attackable)
}

func (a *ActionBasicAttack) MouseOver(bx,by float64) {
}

func (a *ActionBasicAttack) MouseClick(bx,by float64) ActionCommit {
  for i := range a.targets {
    if int(bx) == a.targets[i].Pos.Xi() && int(by) == a.targets[i].Pos.Yi() {
      a.mark = a.targets[i]
      return StandardAction
    }
  }
  return NoAction
}

func (a *ActionBasicAttack) aiDoAttack(mark *Entity) bool {
  if a.Range < mark.Pos.Dist(a.Ent.Pos) { return false }
  if a.Ent.CurAp() < a.Cost { return false }
  a.mark = mark
  return true
}

func (a *ActionBasicAttack) Maintain(dt int64) MaintenanceStatus {
  if a.mark == nil || a.Ent.CurAp() < a.Cost {
    a.Cancel()
    return Complete
  }
  a.Ent.SpendAp(a.Cost)

  if a.Melee != 0 {
    a.Ent.s.Command("melee")
  } else {
    a.Ent.s.Command("ranged")
  }


  attack := a.Power + a.Ent.CurAttack() + ((Dice("5d5") - 2) / 3 - 4)
  defense := a.mark.CurDefense()

  a.mark.s.Command("defend")
  if attack <= defense {
    a.mark.s.Command("undamaged")
  } else {
    a.mark.DoDamage(attack - defense)
    if a.mark.CurHealth() <= 0 {
      a.mark.s.Command("killed")
    } else {
      a.mark.s.Command("damaged")
    }
  }

  a.Ent.turnToFace(a.mark.Pos)

  a.Cancel()
  return Complete
}
