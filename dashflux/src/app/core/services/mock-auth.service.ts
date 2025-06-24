import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { User } from '../store/models';

export const MOCK_USER = {
  email: 'asdf@asdf.com',
  password: 'asdf'
};

export const EXISTING_USER = {
  email: '1234@1234.com',
  password: '1234'
};

export const INVALID_CREDENTIALS_USER = {
  email: 'pera@pera.com',
  password: 'pera'
};

@Injectable()
export class MockAuthService {
  public signup(user: User): Observable<User> {
    if (user.email === MOCK_USER.email) {
      return Observable.of(user).delay(5000);
    }

    if (user.email === EXISTING_USER.email) {
      return Observable.throw(new Error('User with email already exists.'));
    }

    return Observable.throw(new Error('Failed connecting to server.'));
  }

  public login(user: User): Observable<User> {
    if (user.email === MOCK_USER.email) {
      return Observable.of(user).delay(1000);
    }

    if (user.email === INVALID_CREDENTIALS_USER.email) {
      return Observable.throw(new Error('Invalid credentials'));
    }

    return Observable.throw(new Error('Cannot connect to server'));
  }
}
