package gui

import (
  "glop/gin"
)

// The GUI is handled in four steps:
// 1. Handle Events
//   As event groups are received from gin they are passed, one by one, towards whatever widget
//   is in focus.  Each widget that these events are passed through may decide use the events,
//   for example, a table widget that receives an event saying that the tab key was pressed may
//   consume this event and change focus from one widget it contains to another.
//
// 2. Thinking
//   Widget.Think() is called for all widgets only after events are processed.  This gives
//   widgets a chance to take focus based on input other that event groups that are passed
//   around in step 1.  Care must be taken to ensure that widgets are not competing for focus.
//   Widgets should figure out their size during Think().  Think is called on the leaf nodes
//   before the internal nodes so a widget can query its children for their most up-to-date size
//   during Think().
//
// 3. Draw
//   Widgets are recursively called to draw themselves.
//   TODO: Figure out how to set the scissor box for all widgets to enforce the size their parent
//         suggests for them

// Uninstalled widgets will not Think(), and cannot take focus


type Widget interface {
  // Called once per frame.  The widget is responsible for updating its new size here.  Think is
  // called on all of a Widget's installed children before being called on that widget, so it may
  // query its children for their most up-to-date sizes.
  // A widget may return true from this method to request focus.
  Think(ms int64, has_focus bool) bool

  // Returns the width and height that this widget wants to render itself.
  Size() (int, int)

  // Draws the widget, never going outside of the specified region.
  Draw(x,y,dx,dy int)

  // This method is called for every widget in the path from the root to the widget with focus.
  // Every widget along the way has a chance to react to the event group before it gets passed
  // along.
  // consume: If this is true the event group will not be passed to any more nodes.
  // give: If this is true focus will be given to the node specified by target, if target is nil
  //       then focus will be popped.
  // target: If give is set this node will receive focus and regardless of consume the event
  // group will not be passed to any more nodes.
  HandleEventGroup(gin.EventGroup) (consume bool, give bool, target *Node)

  // Any time a new widget is installed in a node, that node's widget will have this method
  // called with the new widget so that it can keep track of it if it wants.  Params is
  // passed directly from InstallWidget to here so it may contain any information.
  AddChild(w Widget, params interface{})

  // Like AddChild, but called when a widget is uninstalled.
  RemoveChild(w Widget)
}

type Node struct {
  widget   Widget
  parent   *Node
  children []*Node
}

// Returns an array of all of the nodes from the root to this node, in that order.
func (n *Node) pathFromRoot() []*Node {
  var path []*Node
  for p := n; p != nil; p = p.parent {
    path = append(path, p)
  }
  for i := 0; i < len(path)/2; i++ {
    path[i],path[len(path)-i-1] = path[len(path)-i-1],path[i]
  }
  return path
}

// Calls Think on all widgets in this node and its descendants.  Think is called first on
// the leaves, then on the internal nodes.
func (n *Node) think(ms int64, focus *Focus) {
  for _,child := range n.children {
    child.think(ms, focus)
  }
  // TODO: perhaps handle the case where multiple widgets try to take focus here?
  //  maybe it should be an error, or maybe just pick one but not actually let it happened
  //  until after everything has Think()ed?
  if n.widget.Think(ms, n == focus.top()) {
    focus.Take(n)
  }
}

func (n *Node) InstallWidget(w Widget, params interface{}) *Node {
  kid := new(Node)
  kid.parent = n
  kid.widget = w
  n.children = append(n.children, kid)
  n.widget.AddChild(w, params)
  return kid
}
func (n *Node) UninstallWidget(w Widget) {
  cur := 0
  for i := range n.children {
    n.children[cur] = n.children[i]
    if n.children[i].widget == w {
      n.children[i].parent = nil
    } else {
      cur++
    }
  }
  n.children = n.children[0 : cur]
  n.widget.RemoveChild(w)
}

// A Focus object tracks what widget has focus.  The widget with focus is the one that events
// will be directed to.  Every incoming EventGroup will be sent first to the root widget, then
// it will pass it to a child widget and so on until it reaches the widget with focus.  There
// are cases when a widget will want to send events elsewhere, for example consider a table with
// two text boxes, A and B, A has focus, B does not.  If the user clicks on B the table widget
// will want to notify B that it should take focus, so it calls focus.Give(B).  This will result
// in B.TookFocus(event_group) being called, so it knows that it has focus and the event that
// made this happen.
type Focus struct {
  nodes []*Node
}

func (f *Focus) top() *Node {
  if len(f.nodes) == 0 {
    return nil
  }
  return f.nodes[0]
}

// Whatever widget currently has focus loses it, and the widget passed to this function gains it.
func (f *Focus) Take(n *Node) {
  if len(f.nodes) == 0 {
    f.nodes = append(f.nodes, nil)
  }
  f.nodes[len(f.nodes)-1] = n
}

// Whatever widget has focus now loses it, but will regain it when Focus.Pop() is called
func (f *Focus) Push(n *Node) {
  f.nodes = append(f.nodes, n)
}

func (f *Focus) Pop() {
  if len(f.nodes) == 0 {
    panic("Cannot pop an empty Focus stack")
  }
  f.nodes = f.nodes[0 : len(f.nodes)-1]
}
