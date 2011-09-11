package system

import (
  "glop/gin"
)

type Window uintptr

type System interface {
  // Call after runtime.LockOSThread(), *NOT* in an init function
  Startup()

  // Call System.Think() every frame
  Think()

  CreateWindow(x,y,width,height int) Window
  // TODO: implement this:
  // DestroyWindow(Window)

  // Self-explanitory getters
  GetWindowPosition(window Window) (int,int)
  GetWindowSize(window Window) (int,int)

  SwapBuffers(window Window)
  GetInputEvents() []gin.EventGroup

  Input() *gin.Input

  // These probably shouldn't be here, probably always want to do the Think() approach
//  Run()
//  Quit()
}

// This is the interface implemented by any operating system that supports
// glop.  The glop/gos package for that OS should export a function called
// GetSystemInterface() which takes no parameters and returns an object that
// implements the system.Os interface.
type Os interface {
  // This is properly called after runtime.LockOSThread(), not in an init function
  Startup()

  // Think() is called on a regular basis and always from main thread.
  Think()

  // Create a window with the appropriate dimensions and bind an OpenGl contxt to it.
  // Currently glop only supports a single window, but this function could be called
  // more than once since a window could be destroyed so it can be recreated at different
  // dimensions or in full sreen mode.
  CreateWindow(x,y,width,height int) Window

  // TODO: implement this:
  // DestroyWindow(Window)

  // Self-explanitory getters
  GetWindowPosition(window Window) (int,int)
  GetWindowSize(window Window) (int,int)

  // Swap the OpenGl buffers on this window
  SwapBuffers(window Window)

  // Returns all of the events in the order that they happened since the last call to
  // this function.  The events do not have to be in order according to KeyEvent.Timestamp,
  // but they will be sorted according to this value.  The timestamp returned is the event
  // horizon, no future events will have a timestamp less than or equal to it.
  GetInputEvents() ([]gin.OsEvent, int64)

  // These probably shouldn't be here, probably always want to do the Think() approach
//  Run()
//  Quit()
}

type sysObj struct {
  os     Os
  input  *gin.Input
  events []gin.EventGroup
}
func Make(os Os) System {
  return &sysObj{
    os : os,
    input : gin.Make(),
  }
}
func (sys *sysObj) Startup() {
  sys.os.Startup()
}
func (sys *sysObj) Think() {
  sys.os.Think()
  events,_ := sys.os.GetInputEvents()
  sys.events = sys.input.Think(-1, false, events)
}
func (sys *sysObj) Input() *gin.Input {
  return sys.input
}
func (sys *sysObj) CreateWindow(x,y,width,height int) Window {
  return sys.os.CreateWindow(x, y, width, height)
}
func (sys *sysObj) GetWindowPosition(window Window) (int,int) {
  return sys.os.GetWindowPosition(window)
}
func (sys *sysObj) GetWindowSize(window Window) (int,int) {
  return sys.os.GetWindowSize(window)
}
func (sys *sysObj) SwapBuffers(window Window) {
  sys.os.SwapBuffers(window)
}
func (sys *sysObj) GetInputEvents() []gin.EventGroup {
  return sys.events
}
