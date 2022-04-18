import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from 'environments/environment';
import { Observable, of } from 'rxjs';

@Injectable()
export class AuthService {
    constructor(private http: HttpClient) {}

    getAuth$(): Observable<{}> {
        return of({});
    }


    /**
     * Login
     * @param credentials 
     * @returns 
     */
    login(credentials: {Email: String, Password: String}): Observable<any>{
      return this.http.post(`${environment.apiUrl}/login`, credentials);
    }

    register(user:any): Observable<any>{
      return this.http.post(`${environment.apiUrl}/register`, user);
    }
}
