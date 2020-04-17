// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { Directive, ElementRef, Renderer2, Input, OnChanges, SimpleChange } from '@angular/core';
@Directive({
  selector: '[loading-display]' // Attribute selector
})

export class LoadingDisplayDirective implements OnChanges {
  @Input('loading-display') loadingDisplay: any;

  constructor(
    public element: ElementRef,
    public renderer: Renderer2
  ) { }

  ngOnInit() {
    this.renderer.addClass(this.element.nativeElement, 'loading-display-initial');
  }

  ngOnChanges(changes: { [propertyName: string]: SimpleChange }) {
    if (changes['loadingDisplay'] && !this.loadingDisplay) {
      this.renderer.addClass(this.element.nativeElement, 'loading-display-loaded');
    }
  }
}
