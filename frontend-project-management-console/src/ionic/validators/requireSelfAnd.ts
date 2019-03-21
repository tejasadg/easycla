import {
    FormGroup,
    FormControl,
    ValidationErrors,
    ValidatorFn,
    Validators,
} from '@angular/forms';

export class ExtraValidators {

  static requireSelfOr(control: FormControl, otherControl : string): any {
    if (control.parent === undefined) {
      return null
    } else {
      if (Validators.required(control) === null){ // Valid
        control.parent.controls[otherControl].updateValueAndValidity()
        return null
      } else { // Not valid
        if (Validators.required(control.parent.controls[otherControl]) === null){ // valid
          return null
        } else { // valid
          return {"required": true}
        }
      }
    }
  }

}