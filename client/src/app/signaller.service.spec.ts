import { TestBed } from '@angular/core/testing';

import { SignallerService } from './signaller.service';

describe('SignallerService', () => {
  let service: SignallerService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SignallerService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
