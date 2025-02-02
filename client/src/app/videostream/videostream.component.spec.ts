import { ComponentFixture, TestBed } from '@angular/core/testing';

import { VideostreamComponent } from './videostream.component';

describe('VideostreamComponent', () => {
  let component: VideostreamComponent;
  let fixture: ComponentFixture<VideostreamComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [VideostreamComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(VideostreamComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
