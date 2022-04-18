import { Injectable } from '@angular/core';
import { CanActivate, Router } from '@angular/router';
import { Observable, of } from 'rxjs';

@Injectable()
export class AuthGuard implements CanActivate {
    constructor(private router: Router){}
    canActivate(): Observable<boolean> {
       const login = localStorage.getItem('login');
       if(login) {
        return of(true);
       }
       this.router.navigate(['auth/login']);
       return of(false);
    }
}
