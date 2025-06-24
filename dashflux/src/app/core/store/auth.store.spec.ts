import { HttpClientModule } from '@angular/common/http';
import { inject, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { RouterTestingModule } from '@angular/router/testing';
import { Observable } from 'rxjs/Observable';

import { AuthenticationService } from '../services/auth/authentication.service';
import { TokenStorage } from '../services/auth/token-storage.service';
import { ChannelsService } from '../services/channels/channels.service';
import { ClientsService } from '../services/clients/clients.service';
import { AuthStore } from './auth.store';
import { UiStore } from './ui.store';

describe('AuthStore', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [
                HttpClientModule,
                RouterTestingModule.withRoutes([])
            ],
            providers: [
                AuthStore,
                UiStore,
                TokenStorage,
                AuthenticationService,
                ClientsService,
                ChannelsService,
            ]
        });
    });

    it('should be created', inject([AuthStore], (authStore: AuthStore) => {
        expect(authStore).toBeTruthy();
    }));

    describe('login', () => {
        const user = {
            email: 'user@user.com',
            password: 'userPassword',
        };

        it('should set the loading flag to true before calling service', inject([AuthStore, UiStore, AuthenticationService],
            (authStore: AuthStore, uiStore: UiStore, authService: AuthenticationService) => {
                const spy = spyOn(authService, 'login').and.returnValue({ subscribe: () => { } });

                authStore.login(user);

                expect(uiStore.loading).toBeTruthy();
            }));

        it('should set the isAuthenticated flag to true when successfully authenticated', inject([AuthStore, AuthenticationService, Router],
            (authStore: AuthStore, authService: AuthenticationService, router: Router) => {
                const spy = spyOn(authService, 'login').and.returnValue(Observable.of(true));
                const routerSpy = spyOn(router, 'navigate').and.stub();

                authStore.login(user);

                expect(authStore.isAuthenticated).toBeTruthy();
            }));


        it('should navigate to /clients when successfully authenticated', inject([AuthStore, AuthenticationService, Router],
            (authStore: AuthStore, authService: AuthenticationService, router: Router) => {
                const spy = spyOn(authService, 'login').and.returnValue(Observable.of(true));
                const routerSpy = spyOn(router, 'navigate').and.stub();

                authStore.login(user);

                expect(routerSpy).toHaveBeenCalledWith(['/clients']);
            }));

        it('should set the loading flag to false when successfully authenticated',
            inject([AuthStore, UiStore, AuthenticationService, Router],
            (authStore: AuthStore, uiStore: UiStore, authService: AuthenticationService, router: Router) => {
                const spy = spyOn(authService, 'login').and.returnValue(Observable.of(true));
                const routerSpy = spyOn(router, 'navigate').and.stub();

                authStore.login(user);

                expect(uiStore.loading).toBeFalsy();
            }));

        it('should set the loading flag to false when authentication failed', inject([AuthStore, UiStore, AuthenticationService],
            (authStore: AuthStore, uiStore: UiStore, authService: AuthenticationService) => {
                const spy = spyOn(authService, 'login').and.returnValue(Observable.throw({ status: 403 }));

                authStore.login(user);

                expect(uiStore.loading).toBeFalsy();
            }));

        it('should set the authError to authentication error', inject([AuthStore, AuthenticationService],
            (authStore: AuthStore, authService: AuthenticationService) => {
                const spy = spyOn(authService, 'login').and.returnValue(Observable.throw('Auth failed'));

                authStore.login(user);

                expect(authStore.authError).toEqual('Auth failed');
            }));
    });

    describe('signup', () => {
        const user = {
            email: 'user@user.com',
            password: 'userPassword',
        };

        it('should set the loading flag to true before calling service', inject([AuthStore, UiStore, AuthenticationService],
            (authStore: AuthStore, uiStore: UiStore, authService: AuthenticationService) => {
                const spy = spyOn(authService, 'signup').and.returnValue({ subscribe: () => { } });

                authStore.signup(user);

                expect(uiStore.loading).toBeTruthy();
            }));

        it('should call the login when signup successfull', inject([AuthStore, AuthenticationService, Router],
            (authStore: AuthStore, authService: AuthenticationService, router: Router) => {
                const spy = spyOn(authService, 'signup').and.returnValue(Observable.of(true));
                const routerSpy = spyOn(router, 'navigate').and.stub();
                const loginSpy = spyOn(authStore, 'login').and.stub();

                authStore.signup(user);

                expect(loginSpy).toHaveBeenCalled();
            }));

        it('should set the loading flag to false when signup failed', inject([AuthStore, UiStore, AuthenticationService],
            (authStore: AuthStore, uiStore: UiStore, authService: AuthenticationService) => {
                const spy = spyOn(authService, 'signup').and.returnValue(Observable.throw('Signup failed'));

                authStore.signup(user);

                expect(uiStore.loading).toBeFalsy();
            }));

        it('should set the authError to signup error', inject([AuthStore, AuthenticationService],
            (authStore: AuthStore, authService: AuthenticationService) => {
                const spy = spyOn(authService, 'signup').and.returnValue(Observable.throw('Signup failed'));

                authStore.signup(user);

                expect(authStore.authError).toEqual('Signup failed');
            }));
    });

    describe('logout', () => {
        it('should call the authentication service logout', inject([AuthStore, AuthenticationService, Router],
            (authStore: AuthStore, authService: AuthenticationService, router: Router) => {
                const spy = spyOn(authService, 'logout');
                const routerSpy = spyOn(router, 'navigate').and.stub();

                authStore.logout();

                expect(spy).toHaveBeenCalled();
            }));

        it('should set the isAuthenticated flag to false', inject([AuthStore, AuthenticationService, Router],
            (authStore: AuthStore, authService: AuthenticationService, router: Router) => {
                const spy = spyOn(authService, 'logout');
                const routerSpy = spyOn(router, 'navigate').and.stub();

                authStore.logout();

                expect(authStore.isAuthenticated).toBeFalsy();
            }));

        it('should navigate to /login', inject([AuthStore, AuthenticationService, Router],
            (authStore: AuthStore, authService: AuthenticationService, router: Router) => {
                const spy = spyOn(authService, 'logout');
                const routerSpy = spyOn(router, 'navigate').and.stub();

                authStore.logout();

                expect(routerSpy).toHaveBeenCalledWith(['/login']);
            }));
    });
});
