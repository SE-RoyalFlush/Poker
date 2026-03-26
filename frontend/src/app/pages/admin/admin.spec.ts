import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';

import { AdminComponent } from './admin';
import { AdminService } from '../../core/services';

describe('AdminComponent', () => {
  let fixture: ComponentFixture<AdminComponent>;
  let component: AdminComponent;
  let adminServiceSpy: jasmine.SpyObj<AdminService>;

  beforeEach(async () => {
    adminServiceSpy = jasmine.createSpyObj<AdminService>('AdminService', ['getUsers', 'deleteUser']);

    adminServiceSpy.getUsers.and.returnValue(
      of([
        { id: 1, username: 'alice' },
        { id: 2, username: 'bob' },
      ])
    );
    adminServiceSpy.deleteUser.and.returnValue(of(void 0));

    await TestBed.configureTestingModule({
      imports: [AdminComponent],
      providers: [{ provide: AdminService, useValue: adminServiceSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(AdminComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should load users with admin credentials', () => {
    component.credentialsForm.setValue({ username: 'admin', password: 'admin' });
    component.onAdminLogin();

    expect(adminServiceSpy.getUsers).toHaveBeenCalledWith('admin', 'admin');
    expect(component.users.length).toBe(2);
  });

  it('should delete user and remove it from list', () => {
    component.credentialsForm.setValue({ username: 'admin', password: 'admin' });
    component.onAdminLogin();
    component.deleteUser(1);

    expect(adminServiceSpy.deleteUser).toHaveBeenCalledWith('admin', 'admin', 1);
    expect(component.users.some((user) => user.id === 1)).toBeFalse();
  });

  it('should show error on invalid admin login', () => {
    adminServiceSpy.getUsers.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 401, statusText: 'Unauthorized' }))
    );

    component.credentialsForm.setValue({ username: 'admin', password: 'admin' });
    component.onAdminLogin();

    expect(component.loginError).toContain('Invalid admin credentials.');
    expect(component.users.length).toBe(0);
  });
});
