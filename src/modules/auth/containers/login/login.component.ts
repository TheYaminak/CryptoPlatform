import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '@modules/auth/services';
import { delay } from 'rxjs/operators';

@Component({
    selector: 'sb-login',
    templateUrl: './login.component.html',
    styleUrls: ['login.component.scss'],
})
export class LoginComponent implements OnInit {

    public loginForm: FormGroup | any;
    public errorMessage: boolean = false;
    public laoding: boolean = false;
    constructor(private authService:AuthService, private fb: FormBuilder, private router: Router) {}
    ngOnInit() {
        this.buildForm();
    }

    buildForm(): FormGroup {
        return this.loginForm = this.fb.group({
          Email: [null, Validators.required],
          Password: [null, Validators.required]
        });
      }

    login(){
        if (this.loginForm.invalid) return;
        this.laoding = true;
        this.errorMessage = false;
       this.authService.login(this.loginForm.value).pipe(delay(600)).subscribe(
           ({id}) => {
                this.laoding = false
               if(id!=0){
                 localStorage.setItem('login', id);
                 this.router.navigate(['dashboard']);
                 return;
               } 
                this.errorMessage = true;
           }
       );
    }
}
