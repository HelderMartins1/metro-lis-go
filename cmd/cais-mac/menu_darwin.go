package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

// webview opens a bare window with no menu bar, and without one macOS has no
// Cmd+Q, no copy and paste, and no Cmd+R. The actions travel the responder
// chain, so the WKWebView answers copy:, paste: and reload: itself.
static NSMenuItem *add(NSMenu *menu, NSString *title, SEL action, NSString *key) {
	return [menu addItemWithTitle:title action:action keyEquivalent:key];
}

static NSMenuItem *submenu(NSMenu *bar, NSMenu *menu) {
	NSMenuItem *item = [[NSMenuItem alloc] init];
	item.submenu = menu;
	[bar addItem:item];
	return item;
}

static void setupMenu(void) {
	NSMenu *bar = [[NSMenu alloc] init];

	NSMenu *app = [[NSMenu alloc] initWithTitle:@"Cais"];
	add(app, @"Acerca do Cais", @selector(orderFrontStandardAboutPanel:), @"");
	[app addItem:[NSMenuItem separatorItem]];
	add(app, @"Ocultar o Cais", @selector(hide:), @"h");
	NSMenuItem *others = add(app, @"Ocultar outros", @selector(hideOtherApplications:), @"h");
	others.keyEquivalentModifierMask = NSEventModifierFlagOption | NSEventModifierFlagCommand;
	add(app, @"Mostrar tudo", @selector(unhideAllApplications:), @"");
	[app addItem:[NSMenuItem separatorItem]];
	add(app, @"Sair do Cais", @selector(terminate:), @"q");
	submenu(bar, app);

	NSMenu *edit = [[NSMenu alloc] initWithTitle:@"Edição"];
	add(edit, @"Desfazer", @selector(undo:), @"z");
	NSMenuItem *redo = add(edit, @"Refazer", @selector(redo:), @"z");
	redo.keyEquivalentModifierMask = NSEventModifierFlagShift | NSEventModifierFlagCommand;
	[edit addItem:[NSMenuItem separatorItem]];
	add(edit, @"Cortar", @selector(cut:), @"x");
	add(edit, @"Copiar", @selector(copy:), @"c");
	add(edit, @"Colar", @selector(paste:), @"v");
	add(edit, @"Selecionar tudo", @selector(selectAll:), @"a");
	submenu(bar, edit);

	NSMenu *view = [[NSMenu alloc] initWithTitle:@"Visualização"];
	add(view, @"Recarregar", @selector(reload:), @"r");
	NSMenuItem *full = add(view, @"Ecrã completo", @selector(toggleFullScreen:), @"f");
	full.keyEquivalentModifierMask = NSEventModifierFlagControl | NSEventModifierFlagCommand;
	submenu(bar, view);

	NSMenu *window = [[NSMenu alloc] initWithTitle:@"Janela"];
	add(window, @"Minimizar", @selector(performMiniaturize:), @"m");
	add(window, @"Zoom", @selector(performZoom:), @"");
	add(window, @"Fechar", @selector(performClose:), @"w");
	submenu(bar, window);
	[NSApp setWindowsMenu:window];

	[NSApp setMainMenu:bar];
}
*/
import "C"

func setupMenu() {
	C.setupMenu()
}
