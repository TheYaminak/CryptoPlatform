import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '@modules/auth/services';
import { delay } from 'rxjs/operators';

@Component({
    selector: 'sb-register',
    templateUrl: './register.component.html',
    styleUrls: ['register.component.scss'],
})
export class RegisterComponent implements OnInit {
    public registerForm: FormGroup | any;
    public errorMessage: boolean = false;
    public laoding: boolean = false;
    constructor(private authService: AuthService, private fb: FormBuilder, private router: Router) {}
    
    ngOnInit() {
        this.buildForm();
    }

    buildForm(): FormGroup {
        return this.registerForm = this.fb.group({
          Email: [null, Validators.required],
          Password: [null, Validators.required],
          Name: [null, Validators.required],
          LastName: [null, Validators.required]
        });
    }

    register(){
        console.log('aquiiii');
        
        if (this.registerForm.invalid) return;
        this.laoding = true;
        this.errorMessage = false;
       this.authService.register(this.registerForm.value).subscribe(
           ({id}) => {
               console.log(id);
               
                this.laoding = false
               if(id!=0){
                localStorage.setItem('login', id);
                 this.router.navigate(['payments']);
                 return;
               } 
                this.errorMessage = true;
           },
           error => console.log(error)
       );
    }
}
