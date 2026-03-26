import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { AdminService } from './admin.service';

describe('AdminService', () => {
  let service: AdminService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [AdminService],
    });

    service = TestBed.inject(AdminService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should request admin users with basic auth header', () => {
    service.getUsers('admin', 'admin').subscribe((users) => {
      expect(users.length).toBe(1);
      expect(users[0].username).toBe('alice');
    });

    const req = httpMock.expectOne('http://localhost:8080/api/admin/users');
    expect(req.request.method).toBe('GET');
    expect(req.request.withCredentials).toBeTrue();
    expect(req.request.headers.get('Authorization')).toBe('Basic YWRtaW46YWRtaW4=');

    req.flush([{ id: 1, username: 'alice' }]);
  });

  it('should normalize wrapped users payload and legacy key casing', () => {
    service.getUsers('admin', 'admin').subscribe((users) => {
      expect(users.length).toBe(2);
      expect(users[0]).toEqual({ id: 3, username: 'carol' });
      expect(users[1]).toEqual({ id: 4, username: 'dave' });
    });

    const req = httpMock.expectOne('http://localhost:8080/api/admin/users');
    req.flush({
      users: [
        { ID: 3, Username: 'carol' },
        { id: 4, username: 'dave' },
      ],
    });
  });

  it('should delete user with basic auth header', () => {
    service.deleteUser('admin', 'admin', 7).subscribe();

    const req = httpMock.expectOne('http://localhost:8080/api/admin/users/7');
    expect(req.request.method).toBe('DELETE');
    expect(req.request.withCredentials).toBeTrue();
    expect(req.request.headers.get('Authorization')).toBe('Basic YWRtaW46YWRtaW4=');

    req.flush(null);
  });
});
